import React from "react";
// material-ui
import {
    Card,
    CardHeader,
    CardContent,
    Divider,
    Typography,
    Button,
    Collapse,
} from "@material-ui/core";
import {
    ExpandMore as ExpandMoreIcon,
    ExpandLess as ExpandLessIcon,
} from "@material-ui/icons";
// styles
import styles from "./QaCard.module.scss";

const TextCard = (): JSX.Element => {

    const [expanded, setExpanded] = React.useState(false);
    const handleExpandClick = () => {
        setExpanded(!expanded);
    };

    return (
        <div >
            <Card>
                <CardHeader
                    title={<Typography variant="h5">TextCard</Typography>}
                    subheader="テキストカード"

                />
                <Divider/>
                <CardContent>
                    <Collapse in={!expanded} timeout="auto" unmountOnExit>
                        <Typography variant="body1">
                            100文字は開かずにみれるようにしよう！！
                        </Typography>
                    </Collapse>
                    <Collapse in={expanded} timeout="auto" unmountOnExit>
                        <Typography variant="body1">
                            ここには全行表示してあげよう！！！！！！！！！！！！！
                            ！！！！！！！！ 自由記入欄
                        </Typography>
                    </Collapse>
                </CardContent>
                <div >
                    <Button
                        fullWidth
                        onClick={handleExpandClick}
                        startIcon={expanded ? <ExpandLessIcon/> : <ExpandMoreIcon/>}

                    >
                        {expanded ? "LESS" : "MORE"}
                    </Button>
                </div>
            </Card>
        </div>
    );
};

export default TextCard;